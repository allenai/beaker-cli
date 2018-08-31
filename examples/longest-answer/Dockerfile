# javascript example


# https://www.javascriptjanuary.com/blog/building-your-first-node-app-using-docker
# docker build -t node-docker .
# docker run --rm node-docker
# or for test
# docker run --rm -e TESTING=true node-docker

# Specifies the base image we're extending
FROM node:9

# Create the results directory we will place our predictions
# Note: /results/predictions.scv is the default file location leaderboard will look for results
RUN mkdir /results

# Specify the "working directory" for the rest of the Dockerfile
WORKDIR /

# Install packages using NPM 5 (bundled with the node:9 image)
COPY ./package.json /package.json
COPY ./package-lock.json /package-lock.json
RUN npm install --silent

# Add application code
COPY ./model.js /model.js

# When testing on local machine, copy the dev dataset
# Note: ./swag.csv is the default file location leaderboard will put the dataset
COPY ./test-data/swag-dev-head-200.csv /swag-dev-head-200.csv

# Run our program
CMD ["node", "model.js"]
