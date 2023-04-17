import axios from 'axios';

export function postLL1(url, body) {
  const headers = {
    'Content-Type': 'application/json'
  };
  return axios.post(url, body, { headers })
    .then((response) => {
      return response.data;
    })
    .catch((error) => {
      throw error;
    });
}