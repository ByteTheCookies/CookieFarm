import { NextRequest, NextResponse } from 'next/server';
import isAuthenticated from '@/lib/auth';
import { BACKEND_URL } from './lib/constants';
import axios from 'axios';

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const isAuth = await isAuthenticated(request);
  if (pathname != '/login' && !isAuth) {
    console.log('User is not authenticated, redirecting to /login');
    return NextResponse.redirect(new URL('/login', request.url));
  }

  try {
    const res = await axios.get(`${BACKEND_URL}/api/v1/config`, {
      headers: {
        'Content-Type': 'application/json',
        Cookie: request.headers.get('cookie'),
      },
      withCredentials: true,
    });

    if (!res.data.configured && pathname != '/config') {
      return NextResponse.redirect(new URL('/config', request.url));
    }
  } catch (error) {
    console.log(error);
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
    '/((?!api|_next/static|_next/image|_next/data|.*\\..*).*)',
  ],
};
