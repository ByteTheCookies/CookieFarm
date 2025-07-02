import axios from 'axios';

export const fetcher = async (url: string) => {
  try {
    const response = await axios.get(url, {
      headers: {
        'Content-Type': 'application/json',
      },
      withCredentials: true,
    });

    return response.data;
  } catch (error) {
    console.error('Error fetching flags:', error);
    return false;
  }
};
