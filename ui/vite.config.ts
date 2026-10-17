import { defineConfig } from 'vite'
import tsconfigPaths from 'vite-tsconfig-paths'
import { cp } from 'node:fs/promises';
//import typescript from "@rollup/plugin-typescript";
import * as path from "path"


const mode = process.env.NODE_ENV || "production"

const paths = {
    production: `dist/assets/shoelace`,
    development: `shoelace`,
}
const vitePath = `${paths[mode]}`

export default defineConfig({
    resolve: {
        alias: [{
            // Monaco 0.38 imports this vendored module through relative paths.
            find: /^(?:\.{1,2}\/)+dompurify\/dompurify\.js$/,
            replacement: path.resolve(__dirname, 'src/monaco-sanitizer.ts'),
        }],
    },
    plugins: [tsconfigPaths()],
    build: {
        minify: true,
        cssCodeSplit: false,
        outDir: './dist',
        rollupOptions: {
            external: [
                /^node:.*/,
                // "lit",
                // "lit/decorators.js",
            ],

            plugins: [{
                name: 'copy-shoelace-assets',
                async writeBundle() {
                    await cp(
                        path.resolve(__dirname, 'node_modules/@shoelace-style/shoelace/dist/assets'),
                        path.resolve(__dirname, vitePath, 'assets'),
                        { recursive: true },
                    );
                },
            }],
        }
    }
})
