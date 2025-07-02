import axios from 'axios';
import { BACKEND_URL } from './constants';

export default async function isAuthenticated(
  request: Request,
): Promise<boolean> {
  const cookie = request.headers.get('cookie');
  if (!cookie) {
    return false;
  }

  if (!cookie.includes('token=')) {
    return false;
  }

  try {
    const response = await axios.get(`${BACKEND_URL}/api/v1/auth/verify`, {
      headers: {
        Cookie: cookie,
      },
    });

    if (response.status !== 200) {
      return false;
    }
    return true;
  } catch (error) {
    console.error('Error verifying authentication:', error);
    return false;
  }

  return true;
}
