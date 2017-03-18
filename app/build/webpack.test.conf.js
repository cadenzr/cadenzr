var utils = require('./utils')
var webpack = require('webpack')
var merge = require('webpack-merge')
var baseConfig = require('./webpack.base.conf')

var webpackConfig = merge(baseConfig, {
  // use inline sourcemap for karma-sourcemap-loader
  module: {
    rules: [
      {
        test: /\.ts$/,
        enforce: "pre",
        loader: 'istanbul-instrumenter-loader',
        include: path.resolve('src/')
        options: {
          compilerOptions: {
            sourceMap: false,
            inlineSourceMap: true
          }
        }
      },
      utils.styleLoaders()
    ]
  },
  devtool: '#inline-source-map',
  plugins: [
    new webpack.DefinePlugin({
      'process.env': require('../config/test.env')
    })
  ]
})

// no need for app entry during tests
delete webpackConfig.entry

module.exports = webpackConfig

