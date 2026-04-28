import path from 'path';
import vue from '@vitejs/plugin-vue';

export default {
    root: "./resources",
    base: "/resources/views/",
    plugins: [vue()],
    build: {
        assetsDir: 'assets',
        copyPublicDir: false,
        emptyOutDir: true,
        outDir: path.resolve('./public/dist'),
        manifest: "manifest.json",
        rollupOptions: {
            input: [
                "./resources/js/main.ts",
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
