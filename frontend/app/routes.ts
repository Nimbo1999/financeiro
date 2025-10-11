import {
  type RouteConfig,
  index,
  route,
  layout,
  prefix,
} from "@react-router/dev/routes";

export default [
  layout("routes/dashboard/layout.tsx", [
    index("routes/dashboard/home.tsx"),
    route("transactions", "routes/dashboard/transactions.tsx"),
    route("upload", "routes/dashboard/upload.tsx"),
    ...prefix("users", [
      index("routes/dashboard/users/home.tsx"),
      route("create", "routes/dashboard/users/create.tsx"),
    ]),
    route("settings", "routes/dashboard/settings.tsx"),
  ]),

  layout("routes/auth/layout.tsx", [
    route("login", "routes/auth/login.tsx"),
    route("code", "routes/auth/code.tsx"),
    route("refresh", "routes/auth/refresh.tsx"),
  ]),
] satisfies RouteConfig;
