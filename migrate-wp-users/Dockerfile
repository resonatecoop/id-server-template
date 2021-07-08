FROM perl:5.20

RUN cpan install Bundle::DBI \
  cpan install DBD::mysql \
  cpan install DBD::Pg \
  cpan install Dotenv \
  cpan install Switch \
  cpan install UUID::Tiny \
  && mkdir -p /var/www/app

WORKDIR /var/www/app

COPY . /var/www/app

CMD [ "perl", "./migrate_wp_user_to_oauth.pl" ]
