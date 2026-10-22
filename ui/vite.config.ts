import { defineConfig } from 'vite'
import { cp } from 'node:fs/promises';
import * as path from "path"


const mode = process.env.NODE_ENV || "production"

const paths = {
    production: `dist/assets/shoelace`,
    development: `shoelace`,
}
const vitePath = `${paths[mode]}`

export default defineConfig({
    resolve: {
        tsconfigPaths: true,
        alias: [{
            // Monaco no longer exports its aggregate stylesheet.
            find: 'monaco-editor/min/vs/editor/editor.main.css?inline',
            replacement: path.resolve(__dirname, 'node_modules/monaco-editor/min/vs/editor/editor.main.css') + '?inline',
        }, {
            // Monaco imports its vendored sanitizer through relative paths.
            find: /^(?:\.{1,2}\/)+dompurify\/dompurify\.js$/,
            replacement: path.resolve(__dirname, 'src/monaco-sanitizer.ts'),
        }],
    },
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
