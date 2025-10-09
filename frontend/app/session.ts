import { createCookieSessionStorage, redirect } from "react-router";
import type { Session } from "~/models/session";
import { env } from "./environment";

class SessionError extends Error {
  constructor(message: string) {
    super(message);
    this.name = SessionError.name;
  }

  static new(message: string) {
    return new SessionError(message);
  }

  public redirect(): never {
    throw redirect("/login");
  }
}

export const { getSession, commitSession, destroySession } =
  createCookieSessionStorage<{ session: Session }>({
    cookie: {
      isSigned: true,
      name: "__session",
      sameSite: "lax",
      path: "/",
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      secrets: [env.SESSION_SECRET],
      maxAge: 60 * 60 * 24 * 7, // 7 days
    },
  });

export const getCurrentSession = async (cookie: string | null) => {
  try {
    const session = await getSession(cookie);
    const sessionObj = session.get("session");
    if (typeof sessionObj === "undefined") {
      throw SessionError.new("No session found");
    }
    return sessionObj;
  } catch (error: unknown) {
    if (error instanceof SessionError) {
      throw error.redirect();
    }
    throw error;
  }
};
