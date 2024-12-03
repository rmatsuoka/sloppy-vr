import basicSsl from "@vitejs/plugin-basic-ssl";

export default {
  // plugins: [
  //   basicSsl({
  //     /** name of certification */
  //     name: "test",
  //     /** custom trust domains */
  //     domains: ["*.custom.com"],
  //     /** custom certification directory */
  //     certDir: "/Users/.../.devServer/cert",
  //   }),
  // ],

  server: {
    proxy: {
      "/socketserver": {
        target: "ws://localhost:8001",
        ws: true,
        rewriteWsOrigin: true,
      },
    },
  },
};
