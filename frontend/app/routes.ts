import {
  type RouteConfig,
  index,
  route,
  layout,
} from "@react-router/dev/routes";

export default [
  layout("routes/dashboard/layout.tsx", [
    index("routes/dashboard/home.tsx"),
    route("transactions", "routes/dashboard/transactions.tsx"),
    route("upload", "routes/dashboard/upload.tsx"),
    route("users", "routes/dashboard/users.tsx"),
    route("settings", "routes/dashboard/settings.tsx"),
  ]),

  layout("routes/auth/layout.tsx", [
    route("login", "routes/auth/login.tsx"),
    route("code", "routes/auth/code.tsx"),
  ]),
] satisfies RouteConfig;
