import basicSsl from "@vitejs/plugin-basic-ssl";

export default {
  host: true,
  plugins: [
    basicSsl({
      /** name of certification */
      name: "test",
      /** custom trust domains */
      domains: ["*.custom.com"],
      /** custom certification directory */
      certDir: "/Users/.../.devServer/cert",
    }),
  ],
};
