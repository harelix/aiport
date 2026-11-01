import {defineConfig} from 'vite';

// Development only (`make dev`): Vite serves the dashboard with live reload on
// 127.0.0.1:5173 and proxies the API to the running Go gateway. Production
// builds ignore `server` and embed the bundle into the Go binary.
const gateway = process.env.AIPORT_DEV_GATEWAY || 'http://127.0.0.1:8787';
const devPort = Number(process.env.AIPORT_DEV_PORT || 5173);
const devOrigins = new Set([`http://127.0.0.1:${devPort}`, `http://localhost:${devPort}`]);

const toGateway = {
  target: gateway,
  changeOrigin: true, // Host becomes the gateway's configured address
  configure(proxy: {on: (event: 'proxyReq', fn: (req: {getHeader(n: string): unknown; setHeader(n: string, v: string): void}) => void) => void}) {
    proxy.on('proxyReq', req => {
      // The gateway accepts only its own Origin. Rewrite it solely for the dev
      // page; any other site's Origin passes through and is rejected as before.
      const origin = req.getHeader('origin');
      if (typeof origin === 'string' && devOrigins.has(origin)) req.setHeader('origin', gateway);
    });
  },
};

export default defineConfig({
  server: {
    host: '127.0.0.1',
    port: devPort,
    strictPort: true,
    proxy: {'/api': toGateway, '/_connect': toGateway},
  },
});
