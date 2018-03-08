let projectConfig = {
    url: {
        data: 'http://data.hasura/v1/query',
    }
}

if (process.env.ENVIRONMENT === 'dev') {
    projectConfig = {
        url: {
            data: 'https://data.' + process.env.CLUSTER_NAME + '.hasura-app.io/v1/query',
        }
    }
}

module.exports = {
  projectConfig
};
