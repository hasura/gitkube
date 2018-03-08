FROM mhart/alpine-node:7.6.0

WORKDIR /src

ADD src/package.json /src/
#install node modules
RUN npm install

# Add app source files
ADD src /src

CMD ["node", "server.js"]
