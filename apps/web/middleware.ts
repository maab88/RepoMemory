import { withAuth } from "next-auth/middleware";

export default withAuth({
  pages: {
    signIn: "/sign-in",
  },
});

export const config = {
  matcher: [
    "/",
    "/organizations/:path*",
    "/onboarding/:path*",
    "/repositories/:path*",
    "/settings/:path*",
    "/integrations/:path*",
    "/search/:path*",
  ],
};
