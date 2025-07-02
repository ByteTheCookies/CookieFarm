// app/api/protocols/route.ts
import axios from 'axios';
import { NextResponse } from 'next/server';
import { BACKEND_URL } from '@/lib/constants';

export async function GET() {
  try {
    const response = await axios.get(`${BACKEND_URL}/api/v1/protocols`);
    const protocolsArray = response.data.protocols;

    if (!Array.isArray(protocolsArray)) {
      console.error('Expected protocols array but got:', protocolsArray);
      return NextResponse.json(
        { error: 'Invalid protocols format from backend' },
        { status: 500 },
      );
    }

    const transformedData = protocolsArray.map((protocol: string) => ({
      value: protocol,
      label: protocol.replace('.so', ''),
    }));

    console.log('Transformed protocols:', transformedData);
    return NextResponse.json(transformedData);
  } catch (error) {
    console.error('Error fetching protocols:', error);
    return NextResponse.json(
      { error: 'Failed to fetch protocols' },
      { status: 500 },
    );
  }
}
