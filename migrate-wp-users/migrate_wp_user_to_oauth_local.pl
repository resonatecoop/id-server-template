#!/usr/bin/perl

use utf8;
use strict;
use DBI;
use Digest::SHA qw(sha256);
use Switch;
use Scalar::Util qw(reftype);
use UUID::Tiny ':std';
use Dotenv -load;

# print $ENV{REMOTE_SSH_HOST};

# my $pid=fork();
# if (($pid == 0) and ($ENV{REMOTE_SSH_HOST} != 'false'))
#  exec { '/usr/bin/ssh' } 'ssh', '-4', '-L','4306:localhost:3306', '-N', $ENV{REMOTE_SSH_HOST};
  # DOES NOT RETURN
#} elsif ($pid > 0) {
#  sleep(1);
#  print STDERR "child: $pid\n";
#} else {
#  print "fork failed\n";
#}

#END { print STDERR "killing ssh tunnel now\n"; kill('KILL',$pid) if $pid != 0 }

my $dsn = "DBI:mysql:database=$ENV{MYSQL_DB_NAME};host=localhost;port=3306";
my $dbh_wp=DBI->connect($dsn, $ENV{MYSQL_DB_USER}, $ENV{MYSQL_DB_PASS},
  {RaiseError=>1, AutoCommit=>0, FetchHashKeyName=>"NAME_lc", mysql_enable_utf8=>1}
) or die "can't connect to WP database: $DBI::errstr";
$dbh_wp->{'mysql_enable_utf8'}=1;
$dbh_wp->do(qq{SET NAMES 'utf8';});

my $sth=$dbh_wp->prepare('select user_email, user_pass, ID from rsntr_users');
$sth->execute or die "can't get users: $DBI::errstr";
my $wp_users=$sth->fetchall_arrayref;

my $dbh_oauth=DBI->connect("dbi:Pg:dbname=$ENV{PSQL_DB_NAME};host=$ENV{PSQL_DB_HOST}",
	$ENV{PSQL_DB_USER},
	$ENV{PSQL_DB_PASS},
	{RaiseError=>0, AutoCommit=>1}
) or die "can't connect to Oauth2 database: $DBI::errstr";
my $sth=$dbh_oauth->prepare('select * from oauth_users');
$sth->execute or die "can't get users: $DBI::errstr";
my $oauth_users_by_email=$sth->fetchall_hashref('username');

my %seen;

foreach my $row (@$wp_users)
{
	my ($user_email,$user_pass,$id)=@$row;
	next if $user_email eq '';
  $user_email=lc($user_email);
	if($seen{$user_email} > 0)
	{
		print STDERR "WARNING: already seen user email address: $user_email: $seen{$user_email}\n";
		$seen{$user_email}+=1;
		next;
	}
	$seen{$user_email}=1;
	if($oauth_users_by_email->{$user_email})
	{
		if($oauth_users_by_email->{$user_email}{'password'} ne $user_pass)
    {
      print STDERR "updating (password): $user_email\n";
		  my $sth=$dbh_oauth->prepare('update oauth_users set password = ? where username = ?');
      $sth->execute($user_pass,$user_email);
    }
	}
	else
	{
		print STDERR "migrating (inserting): $user_email\n";
		my $sth=$dbh_oauth->prepare('insert into oauth_users (id,created_at,role_id,username,password) values (?,current_timestamp,?,?,?)');
		$sth->execute(create_uuid_as_string(UUID_V4),'user',$user_email,$user_pass); # or die "can't migrate wp_user: $user_email: $DBI::errstr";
	}
}

my $sth=$dbh_wp->prepare('select LOWER(user_email) as user_email from rsntr_users');
$sth->execute or die "can't get users: $DBI::errstr";
my $wp_users_by_email=$sth->fetchall_hashref('user_email');

my %seen2;

foreach my $user (keys %$oauth_users_by_email) {
  my $username = $oauth_users_by_email->{$user}->{'username'};
  my $user_id = $oauth_users_by_email->{$user}->{'id'};

  next if $username eq '';
  $username=lc($username);

  if($seen2{$username} > 0)
  {
    print STDERR "WARNING: already seen username: $username: $seen2{$username}\n";
    $seen2{$username}+=1;
    next;
  }

  $seen2{$username}=1;

  if(!$wp_users_by_email->{$username}) {
    print STDERR "should delete: $username\n";
    my $sth=$dbh_oauth->prepare('delete from oauth_refresh_tokens where user_id = ?');
    $sth->execute($user_id);
    my $sth=$dbh_oauth->prepare('delete from oauth_access_tokens where user_id = ?');
    $sth->execute($user_id);
    my $sth=$dbh_oauth->prepare('delete from oauth_users where username = ?');
    $sth->execute($username);
  }
}

$dbh_wp->disconnect();
$dbh_oauth->disconnect();
