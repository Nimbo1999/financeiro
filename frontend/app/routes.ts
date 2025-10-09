import {
  type RouteConfig,
  index,
  route,
  layout,
} from "@react-router/dev/routes";

export default [
  layout("routes/dashboard.layout.tsx", [index("routes/home.tsx")]),

  layout("routes/auth.layout.tsx", [
    route("login", "routes/login.tsx"),
    route("code", "routes/code.tsx"),
  ]),
] satisfies RouteConfig;
