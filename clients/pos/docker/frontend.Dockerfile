# Packages the already-built POS web SPA (dist/) into an nginx static server.
# The publish workflow builds the web app first (no native/Tauri build), so this
# image is just a fast COPY. Serves the SPA on port 8080 with a try_files
# fallback to index.html.
FROM nginx:alpine

COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY dist /usr/share/nginx/html

EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]
