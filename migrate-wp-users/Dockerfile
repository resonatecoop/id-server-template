FROM perl:5.20

RUN useradd --user-group --shell /bin/false go_oauth2_server

RUN cpan install Bundle::DBI \
  cpan install DBD::mysql \
  cpan isntall DBD::Pg \
  cpan install Dotenv \
  cpan install Switch \
  cpan install UUID::Tiny \
  && mkdir -p /home/go_oauth2_server

COPY . /home/go_oauth2_server

USER go_oauth2_server

WORKDIR /home/go_oauth2_server

CMD [ "perl", "./migrate_wp_user_to_oauth.pl" ]
