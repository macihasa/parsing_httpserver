# Step 1: Use the official Node.js image as the base image for the build stage
FROM node:18 AS build

# Step 2: Set the working directory in the container
WORKDIR /app

# Step 3: Copy the package.json and package-lock.json (if available)
COPY package*.json ./

# Step 4: Install the dependencies
RUN npm install

# Step 5: Copy the entire client directory into the container
COPY . .

# Step 6: Build the production version of the Vite app
RUN npm run build

# Step 7: Use the NGINX image as the base image for the final stage
FROM nginx:alpine

# Step 8: Copy the built app from the build stage to the NGINX HTML directory
COPY --from=build /app/dist /usr/share/nginx/html

# Step 9: Expose the port on which the NGINX server listens
EXPOSE 80

# Step 10: Start the NGINX server when the container runs
CMD ["nginx", "-g", "daemon off;"]