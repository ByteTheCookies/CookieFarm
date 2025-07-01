import { NextRequest, NextResponse } from 'next/server';
import isAuthenticated from '@/lib/auth';

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const isAuth = await isAuthenticated(request);
  if (pathname != '/login' && !isAuth) {
    console.log('User is not authenticated, redirecting to /login');
    return NextResponse.redirect(new URL('/login', request.url));
  }

  const response = NextResponse.next();

  return response;
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    '/((?!_next/static|api|_next/image|images|favicon.ico|sitemap.xml|robots.txt).*)',
  ],
};
