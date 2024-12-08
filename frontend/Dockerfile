# Use the official nginx image to serve static files
FROM nginx:alpine

# Copy the frontend files to nginx's html folder
COPY frontend /usr/share/nginx/html

# Expose port 80 to access the website
EXPOSE 80

# Start nginx in the foreground
CMD ["nginx", "-g", "daemon off;"]
