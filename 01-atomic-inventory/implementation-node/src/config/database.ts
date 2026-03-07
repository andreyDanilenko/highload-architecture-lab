import { Pool } from 'pg';
import dotenv from 'dotenv';

dotenv.config();

const dbHost = process.env.DB_HOST || 'localhost';

export const pgPool = new Pool({
  host: dbHost,
  port: parseInt(process.env.DB_PORT || '5432'),
  database: process.env.DB_NAME || 'inventory',
  user: process.env.DB_USER || 'postgres',
  password: process.env.DB_PASSWORD || 'postgres',
  max: 20,
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 2000,
});

pgPool.on('connect', () => {
  console.log(`✅ PostgreSQL connected to ${dbHost}:${process.env.DB_PORT || '5432'}`);
});

pgPool.on('error', (err) => {
  console.error('❌ PostgreSQL error:', err.message);
});
