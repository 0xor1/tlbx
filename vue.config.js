module.exports = {
    devServer: {
      disableHostCheck: true,
      open: true,
      port: 8081,
      proxy: {
        '/api/': {
          target: 'https://task-trees.com',
          ws: true,
          changeOrigin: true
        }
      }
    }
  }