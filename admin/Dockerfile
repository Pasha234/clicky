FROM composer:2 AS composer-base

WORKDIR /app

COPY composer.json composer.lock ./

# The PHP runtime below provides intl. Composer's image does not, so dependency
# resolution must ignore that platform check in both image variants.
FROM composer-base AS development-dependencies
RUN composer install --no-interaction --prefer-dist --optimize-autoloader \
    --ignore-platform-req=ext-intl \
    --no-scripts

FROM composer-base AS production-dependencies
RUN composer install --no-dev --no-interaction --prefer-dist --optimize-autoloader \
    --ignore-platform-req=ext-intl \
    --no-scripts

FROM php:8.4-cli-alpine AS runtime-base

WORKDIR /var/www/html

RUN apk add --no-cache $PHPIZE_DEPS libpq-dev libzip-dev oniguruma-dev icu-dev libxml2-dev \
    && docker-php-ext-install pdo_pgsql mbstring bcmath intl dom \
	&& pecl install redis \
	&& docker-php-ext-enable redis \
    && apk add --no-cache libpq libzip oniguruma icu-libs libxml2 \
    && apk del $PHPIZE_DEPS libpq-dev libzip-dev oniguruma-dev icu-dev libxml2-dev

COPY . .

COPY docker/entrypoint.sh /usr/local/bin/admin-entrypoint
RUN chmod +x /usr/local/bin/admin-entrypoint

ENTRYPOINT ["admin-entrypoint"]

FROM runtime-base AS development

COPY --from=development-dependencies /app/vendor ./vendor

RUN cp .env.example .env \
    && rm -f bootstrap/cache/*.php \
    && php artisan package:discover --ansi \
    && php artisan filament:upgrade --ansi \
    && rm .env \
    && chown -R www-data:www-data storage bootstrap/cache

EXPOSE 8000

FROM runtime-base AS production

COPY --from=production-dependencies /app/vendor ./vendor

RUN rm -f bootstrap/cache/*.php \
    && php artisan package:discover --ansi \
    && php artisan filament:upgrade --ansi \
    && chown -R www-data:www-data storage bootstrap/cache

EXPOSE 8000
