import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  vus: 10,
  duration: '2m',
};

export default function () {
  http.get('http://app:8080/work');
  sleep(1);
}