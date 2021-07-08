<?php
/**
 * Sample configuration for Resonate WordPress
 *
 * @link https://wordpress.org/support/article/editing-wp-config-php/
 *
 * @package WordPress
 */

define( 'DB_NAME', 'resonate_is' );
define( 'DB_USER', 'go_oauth2_server' );
define( 'DB_PASSWORD', '' );
define( 'DB_HOST', 'localhost' );
define( 'DB_CHARSET', 'utf8' );
define( 'DB_COLLATE', '' );
define( 'AUTH_KEY',         '12345678' );
define( 'SECURE_AUTH_KEY',  '12345678' );
define( 'LOGGED_IN_KEY',    '12345678' );
define( 'NONCE_KEY',        '12345678' );
define( 'AUTH_SALT',        '12345678' );
define( 'SECURE_AUTH_SALT', '12345678' );
define( 'LOGGED_IN_SALT',   '12345678' );
define( 'NONCE_SALT',       '12345678' );

$table_prefix = 'rsntr_';

define( 'WP_DEBUG', false );

/* That's all, stop editing! Happy publishing. */

if ( ! defined( 'ABSPATH' ) ) {
	define( 'ABSPATH', __DIR__ . '/' );
}

require_once ABSPATH . 'wp-settings.php';