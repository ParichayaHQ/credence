const path = require('path');

module.exports = {
  mode: process.env.NODE_ENV || 'development',
  entry: {
    'index': './src/preload/index.ts',
    'wallet-api': './src/preload/wallet-api.ts',
    'system-api': './src/preload/system-api.ts',
  },
  target: 'electron-preload',
  devtool: 'source-map',
  resolve: {
    extensions: ['.ts', '.js'],
    alias: {
      '@': path.resolve(__dirname, 'src'),
      '@shared': path.resolve(__dirname, 'src/shared'),
      '@preload': path.resolve(__dirname, 'src/preload'),
    },
  },
  module: {
    rules: [
      {
        test: /\.ts$/,
        exclude: /node_modules/,
        use: {
          loader: 'ts-loader',
          options: {
            configFile: 'tsconfig.preload.json',
          },
        },
      },
    ],
  },
  output: {
    path: path.resolve(__dirname, 'build/preload'),
    filename: '[name].js',
  },
  externals: {
    electron: 'commonjs electron',
  },
  node: {
    __dirname: false,
    __filename: false,
  },
};