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
        //manifest: true,
        minify: true,
        outDir: './dist',
        rollupOptions: {
            external: [
                /^node:.*/,
            ],
            plugins: [copy({
                    targets: [
                        {
                            src: path.resolve(__dirname, 'node_modules/@shoelace-style/shoelace/dist/assets'),
                            dest: path.resolve(__dirname, vitePath)
                        }
                    ],
                hook: 'writeBundle',
                }),
                // typescript({
                //     sourceMap: false,
                //     declaration: true,
                //     outDir: "dist",
                // }),
            ],
        }
    }
})