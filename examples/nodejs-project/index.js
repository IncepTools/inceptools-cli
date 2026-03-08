// examples/nodejs-project/index.js
require('dotenv').config();

const dbUrl = process.env.DATABASE_URL || 'postgres://localhost:5432/mydb';

console.log('--- Cosmic App Starting ---');
console.log(`Target Database: ${dbUrl}`);
console.log('Status: Online');
