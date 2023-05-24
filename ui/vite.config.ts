import { defineConfig } from 'vite'
import tsconfigPaths from 'vite-tsconfig-paths'
import copy from 'rollup-plugin-copy';
import * as path from "path"
import typescript from "@rollup/plugin-typescript";


const mode = process.env.NODE_ENV || "production"

const paths = {
    production: `dist/shoelace`,
    development: `shoelace`,
}
const vitePath = `${paths[mode]}`

export default defineConfig({
    plugins: [tsconfigPaths()],
    build: {
        manifest: true,
        minify: true,
        reportCompressedSize: true,
        lib: {
            entry: path.resolve(__dirname, "src/index.ts"),
            fileName: "index",
            formats: ["es", "cjs"],
        },
        outDir: './dist',
        rollupOptions: {
            external: [
                /^node:.*/,
            ],
            plugins: [
                typescript({
                    sourceMap: false,
                    declaration: true,
                    outDir: "dist",
                }),
                copy({
                    targets: [
                        {
                            src: path.resolve(__dirname, 'node_modules/@shoelace-style/shoelace/dist/assets'),
                            dest: path.resolve(__dirname, vitePath)
                        }
                    ],
                hook: 'writeBundle',
                })
            ],
        }
    }
})