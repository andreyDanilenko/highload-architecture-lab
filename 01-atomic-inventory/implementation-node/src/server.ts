import Fastify from 'fastify';
import dotenv from 'dotenv';
import { pgPool } from './config/database';
import { redisClient, connectRedis } from './config/redis';

dotenv.config();

const fastify = Fastify({
  logger: true,
  trustProxy: true
});

const PORT = parseInt(process.env.PORT || '3000');
const HOST = process.env.HOST || '0.0.0.0';

fastify.get('/health', async (request, reply) => {
  const dbStatus = await pgPool.query('SELECT 1 as connected')
    .then(() => 'ok')
    .catch((err) => {
      console.error('PostgreSQL health check failed:', err.message);
      return 'error';
    });

  const redisStatus = redisClient.isReady ? 'ok' : 'disconnected';

  return { 
    status: 'ok', 
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
    databases: {
      postgres: dbStatus,
      redis: redisStatus
    }
  };
});

const start = async () => {
  try {
    await connectRedis();
    console.log('📦 Redis connected');
    
    await pgPool.query('SELECT 1');
    console.log('📦 PostgreSQL connected');
    
    await fastify.listen({ port: PORT, host: HOST });
    console.log(`🚀 Server running on http://${HOST}:${PORT}`);
    console.log(`📊 Health check: http://${HOST}:${PORT}/health`);
  } catch (err) {
    console.error('❌ Failed to start server:', err);
    process.exit(1);
  }
};

start();

const gracefulShutdown = async (signal: string) => {
  console.log(`\n🛑 Received ${signal}, shutting down gracefully...`);
  
  try {
    await fastify.close();
    console.log('✅ HTTP server closed');
    
    await pgPool.end();
    console.log('✅ PostgreSQL pool closed');
    
    if (redisClient.isReady) {
      await redisClient.quit();
      console.log('✅ Redis connection closed');
    }
    
    console.log('👋 Goodbye!');
    process.exit(0);
  } catch (error) {
    console.error('❌ Error during shutdown:', error);
    process.exit(1);
  }
};

process.on('SIGINT', () => gracefulShutdown('SIGINT'));
process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
