import { defineConfig } from 'vite'
import tsconfigPaths from 'vite-tsconfig-paths'
import copy from 'rollup-plugin-copy';
//import typescript from "@rollup/plugin-typescript";
import * as path from "path"


const mode = process.env.NODE_ENV || "production"

const paths = {
    production: `dist/shoelace`,
    development: `shoelace`,
}
const vitePath = `${paths[mode]}`

export default defineConfig({
    plugins: [tsconfigPaths()],
    build: {
        minify: true,
        cssCodeSplit: false,
        outDir: './dist',
        rollupOptions: {
            external: [
                /^node:.*/,
            ],
            output: {
                inlineDynamicImports: true,
                entryFileNames: '[name].js',   // currently does not work for the legacy bundle
                assetFileNames: '[name].[ext]', // currently does not work for images
            },
            plugins: [copy({
                    targets: [
                        {
                            src: path.resolve(__dirname, 'node_modules/@shoelace-style/shoelace/dist/assets'),
                            dest: path.resolve(__dirname, vitePath)
                        }
                    ],
                hook: 'writeBundle',
                }),
            ],
        }
    }
})