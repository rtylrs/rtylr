# Packages an already-built SPA (dist/<APP>) into an nginx static server.
# The publish workflow builds all apps first, so this image is just a fast COPY.
# Serves the SPA on port 8080 with a try_files fallback to index.html.
FROM nginx:alpine

ARG APP
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY dist/${APP} /usr/share/nginx/html

RUN apk upgrade --no-cache \
  && touch /var/run/nginx.pid \
  && chown -R nginx:nginx /usr/share/nginx/html /var/cache/nginx /var/log/nginx /var/run/nginx.pid

USER nginx

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD wget -q -O /dev/null http://127.0.0.1:8080/ || exit 1
CMD ["nginx", "-g", "daemon off;"]
