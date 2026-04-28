import path from 'path';

export default {
    root: "./resources",
    base: "/resources/views/",
    build: {
        assetsDir: 'assets',
        copyPublicDir: false,
        emptyOutDir: true,
        outDir: path.resolve('./public/dist'),
        manifest: "manifest.json",
        rollupOptions: {
            input: [
                "./resources/js/script.ts",
                "./resources/css/styles.css"
            ],
        },
    },
    server: {
        host: "localhost",
        port: 4001
    },
    optimizeDeps: {
        entries: "*.jet"
    }
}
