import { betterAuth } from "better-auth"

export const auth = betterAuth({
  baseURL: import.meta.env.VITE_APP_URL,
  secret: import.meta.env.VITE_BETTER_AUTH_SECRET,
  database: {
    provider: "sqlite",
    url: ":memory:", // For development only
  },
  emailAndPassword: {
    enabled: false, // We're using Google OAuth only
  },
  socialProviders: {
    google: {
      clientId: import.meta.env.VITE_GOOGLE_CLIENT_ID || "",
      clientSecret: import.meta.env.VITE_GOOGLE_CLIENT_SECRET || "",
    },
  },
  session: {
    cookieCache: {
      enabled: true,
      maxAge: 5 * 60, // Cache for 5 minutes
    },
  },
  advanced: {
    crossSubDomainCookies: {
      enabled: true,
    },
  },
  trustedOrigins: ["http://localhost:8080"],
})

export type Session = typeof auth.$Infer.Session
