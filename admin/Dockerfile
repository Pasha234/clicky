FROM composer:2 AS dependencies

WORKDIR /app

COPY composer.json composer.lock ./
# The final image installs intl. The Composer image does not include it, so
# ignore that one platform check while resolving the locked vendor tree.
RUN composer install --no-dev --no-interaction --prefer-dist --optimize-autoloader \
    --ignore-platform-req=ext-intl \
    --no-scripts

FROM php:8.4-cli-alpine

WORKDIR /var/www/html

RUN apk add --no-cache $PHPIZE_DEPS libpq-dev libzip-dev oniguruma-dev icu-dev \
    && docker-php-ext-install pdo_pgsql mbstring bcmath intl \
    && apk add --no-cache libpq libzip oniguruma icu-libs \
    && apk del $PHPIZE_DEPS libpq-dev libzip-dev oniguruma-dev icu-dev

COPY . .
COPY --from=dependencies /app/vendor ./vendor

RUN rm -f bootstrap/cache/*.php \
    && php artisan package:discover --ansi \
    && php artisan filament:upgrade --ansi \
    && chown -R www-data:www-data storage bootstrap/cache

EXPOSE 8000
