import http from 'k6/http'; export default function() { http.get('http://backend:8080/health'); }
