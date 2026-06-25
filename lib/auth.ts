import type { NextAuthConfig, DefaultSession } from 'next-auth'
import { default as KeycloakProvider } from 'next-auth/providers/keycloak'
import { default as NextAuth } from 'next-auth'
import type { AuthDTO, AuthProfileDTO, RoleDTO } from '@/feat/Auth/dto'
import {
  fetcher,
  isTokenExpired,
  refreshAccessToken,
  shouldRefreshPermissions,
} from '@/feat/Auth/helpers'

declare module 'next-auth' {
  interface Session extends AuthDTO {
    user: DefaultSession['user']
    roles: RoleDTO
  }
}

declare module 'next-auth/jwt' {
  interface JWT extends AuthDTO {
    idToken?: string
  }
}

// Constants
const AUTH_SECRET =
  process.env.AUTH_SECRET || 'development_fallback_secret_must_be_at_least_32_characters'
const KEYCLOAK_ID = process.env.KEYCLOAK_ID || 'keycloak-id'
const KEYCLOAK_SECRET = process.env.KEYCLOAK_SECRET || 'keycloak-secret'
const KEYCLOAK_ISSUER = process.env.KEYCLOAK_ISSUER || 'https://keycloak.example.com/realms/master'
const APP_API_VIDEO_CONFERENCE = process.env.APP_API_VIDEO_CONFERENCE || 'http://localhost:8080'
const SESSION_MAX_AGE = 6 * 60 * 60 // 6 hours in seconds

const AuthOptions: NextAuthConfig = {
  secret: AUTH_SECRET,
  session: {
    maxAge: SESSION_MAX_AGE,
  },
  jwt: {
    maxAge: SESSION_MAX_AGE,
  },
  providers: [
    KeycloakProvider({
      clientId: KEYCLOAK_ID,
      clientSecret: KEYCLOAK_SECRET,
      checks: ['none'],
      issuer: KEYCLOAK_ISSUER,
    }),
  ],
  callbacks: {
    async jwt({ token, account }) {
      // Save account information on initial sign-in
      if (account) {
        token.access_token = account.access_token ?? ''
        token.expires_at = account.expires_at ?? 0
        token.refresh_token = account.refresh_token ?? ''
      }

      // Fetch profile on first request
      if (!token.profile) {
        const { data: profile } = await fetcher<AuthProfileDTO>(
          `${APP_API_VIDEO_CONFERENCE}/api/me`,
          {
            method: 'GET',
            headers: {
              Authorization: `Bearer ${token.access_token}`,
            },
          }
        )
        token.profile = profile
      }

      if (shouldRefreshPermissions(token)) {
        // Refresh permissions
      }

      // Return token it its still valid and got no error
      if (!isTokenExpired(token) && token?.error !== 'RefreshAccessTokenError') {
        return token
      }

      // Otherwise refresh access token
      return await refreshAccessToken(token)
      // return token
    },

    session({ session, token }) {
      session.access_token = token.access_token
      session.refresh_token = token.refresh_token
      session.profile = token.profile
      session.roles = token.profile.role
      session.publicUrl = APP_API_VIDEO_CONFERENCE ?? ''

      return session
    },
  },
}

export const { handlers, auth, signIn, signOut } = NextAuth(AuthOptions)
