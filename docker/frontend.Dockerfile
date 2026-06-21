# Packages an already-built SPA (dist/<APP>) into an nginx static server.
# The publish workflow builds all apps first, so this image is just a fast COPY.
# Serves the SPA on port 8080 with a try_files fallback to index.html.
FROM nginx:alpine

ARG APP
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY dist/${APP} /usr/share/nginx/html

EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]
