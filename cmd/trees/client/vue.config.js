module.exports = {
    devServer: {
      disableHostCheck: true,
      open: true,
      port: 8081,
      proxy: {
        '/api/': {
          target: 'https://task-trees.com',
          ws: true,
          changeOrigin: true,
          onProxyRes: proxyResponse => {
            if (proxyResponse.headers['set-cookie']) {
              const cookies = proxyResponse.headers['set-cookie'].map(cookie =>
                cookie.replace(/; secure/gi, '')
              );
              proxyResponse.headers['set-cookie'] = cookies;
            }
          }
        }
      }
    }
  }