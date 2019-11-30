module.exports = {
  devServer: {
    proxy: {
      '/data': {
        target: 'http://localhost:8000',
        secure: false,
      }
    },
  },
}
