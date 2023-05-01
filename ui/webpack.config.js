const prod = process.env.NODE_ENV === 'production';

const HtmlWebpackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const CssMinimizerPlugin = require("css-minimizer-webpack-plugin");
const TsconfigPathsPlugin = require('tsconfig-paths-webpack-plugin');
const CopyPlugin = require('copy-webpack-plugin');

const path = require("path");

module.exports = {
    mode: prod ? 'production' : 'development',
    entry: {
        bundle: './src/ts/index.ts'
    },
    output: {
        path: path.resolve('build', 'static'),
        filename: '[name].js',
    },
    resolve: {
        plugins: [new TsconfigPathsPlugin()],
        extensions: ['.ts', '.js'],
    },
    module: {
        rules: [
            {
                test: /\.(ts|tsx)$/,
                exclude: /node_modules/,
                resolve: {
                    extensions: ['.ts', '.tsx', '.js', '.json'],
                },
                use: 'ts-loader',
            },
            {
                test: /\.css$/,
                use: [MiniCssExtractPlugin.loader, 'css-loader'],
            },
        ]
    },
    devtool: prod ? undefined : 'source-map',
    devServer: {
        compress: true,
        port: 3000
    },
    optimization: {
        minimizer: [new CssMinimizerPlugin(), '...'],
        minimize: true,
    },
    plugins: [
        new HtmlWebpackPlugin({
            template: 'index.html',
        }),
        new MiniCssExtractPlugin(),
        new CopyPlugin({
            patterns: [
                // Copy Shoelace assets to dist/shoelace
                {
                    from: path.resolve(__dirname, 'node_modules/@shoelace-style/shoelace/dist/assets'),
                    to: path.resolve(__dirname, 'static/shoelace/assets')
                }
            ]
        })
    ]
};