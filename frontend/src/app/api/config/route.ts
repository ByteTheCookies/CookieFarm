import axios from 'axios';
import { NextRequest, NextResponse } from 'next/server';
import { BACKEND_URL } from '@/lib/constants';

export async function POST(request: NextRequest) {
  try {
    const configData = await request.json();
    console.log('Sending config to backend:', configData);

    // Passa i cookie della richiesta al backend
    const cookies = request.headers.get('cookie');

    const response = await axios.post(
      `${BACKEND_URL}/api/v1/config`,
      configData,
      {
        headers: {
          'Content-Type': 'application/json',
          ...(cookies && { Cookie: cookies }),
        },
        withCredentials: true,
      },
    );

    console.log('Backend response:', response.data);
    return NextResponse.json(response.data);
  } catch (error: any) {
    // eslint-disable-line
    console.error('Error sending config:', error);

    if (error.response?.status === 401) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    return NextResponse.json(
      { error: 'Failed to save configuration' },
      { status: 500 },
    );
  }
}
