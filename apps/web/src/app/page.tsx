import { AuthRedirect } from "@/components/auth/auth-redirect";

export default function HomePage() {
  return (
    <AuthRedirect
      toAuthenticated="/dashboard"
      toUnauthenticated="/login"
    />
  );
}
