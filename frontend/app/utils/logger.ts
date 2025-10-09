import pino from "pino";
import { env } from "~/environment";

const logger = pino({
  level: env.LOGGER_LEVEL,
});

export { logger };
