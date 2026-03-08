import pino from "pino";

declare module "fastify" {
  interface FastifyBaseLogger extends pino.Logger {}
}
