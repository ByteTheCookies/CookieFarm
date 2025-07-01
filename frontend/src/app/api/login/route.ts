import axios from 'axios';
import { NextRequest, NextResponse } from 'next/server';
import { BACKEND_URL } from '@/lib/constants';

export async function POST(req: NextRequest) {
  const formData = await req.formData();
  const password = formData.get('password') || '';
  if (!password) {
    return NextResponse.json(
      { error: 'Password is required' },
      { status: 400 },
    );
  }

  const params = new URLSearchParams();
  params.append('username', 'webpage');
  params.append('password', password as string);

  try {
    const response = await axios.post(
      BACKEND_URL + '/api/v1/auth/login',
      params,
      {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
      },
    );

    if (response.status !== 200) {
      return NextResponse.json(
        { error: 'Login failed' },
        { status: response.status },
      );
    }
    if (
      !response.headers['set-cookie'] ||
      response.headers['set-cookie'].length === 0
    ) {
      return NextResponse.json(
        { error: 'No session cookie received' },
        { status: 500 },
      );
    }

    const nextResponse = NextResponse.json({ result: 'Login succefully' });
    nextResponse.headers.set(
      'Set-Cookie',
      response.headers['set-cookie'].join('; '),
    );
    return nextResponse;
  } catch (error) {
    console.error('Error calling external API:', error);
    return NextResponse.json(
      { error: 'Failed to call external API' },
      { status: 500 },
    );
  }
}
