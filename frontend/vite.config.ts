import { reactRouter } from "@react-router/dev/vite";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";
import devtools from "vite-plugin-devtools-json";

export default defineConfig({
  plugins: [reactRouter(), tsconfigPaths(), devtools()],
  ssr: {
    // Workaround for resolving dependencies in the server bundle
    // Without this, the React context will be different between direct import and transitive imports in development environment
    // For more information, see https://github.com/mui/material-ui/issues/45878#issuecomment-2987441663
    optimizeDeps: {
      include: ["@emotion/*", "@mui/*"],
    },
    noExternal: ["@emotion/*", "@mui/*"],
  },
});
