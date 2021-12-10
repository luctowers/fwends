const path = require("path");
const HtmlWebpackPlugin = require("html-webpack-plugin");

module.exports = function (_env, argv) {
  const isProduction = argv.mode === "production";
  const isDevelopment = !isProduction;

  return {
    entry: "./src/index.jsx",
    output: {
      path: path.resolve(__dirname, "dist"),
      publicPath: "/"
    },
    devServer: {
      compress: true,
      allowedHosts: 'all'
    },      
    module: {
      rules: [
        {
          test: /\.jsx?$/,
          exclude: /node_modules/,
          use: [
            "babel-loader"
          ]
        },
        {
          test: /\.css$/,
          use: [
            "style-loader",
            "css-loader",
            "postcss-loader"
          ]
        }
      ]
    },
    plugins: [
      new HtmlWebpackPlugin({
        template: path.resolve(__dirname, "src/index.html"),
        inject: true
      })
    ],
    resolve: {
      extensions: [".js", ".jsx"]
    },
    performance: {
      hints: false,
      maxEntrypointSize: 524288,
      maxAssetSize: 524288
    }
  };
};
