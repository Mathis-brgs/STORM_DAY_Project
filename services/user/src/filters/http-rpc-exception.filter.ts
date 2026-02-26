import { Catch, ArgumentsHost, HttpException } from '@nestjs/common';
import { BaseRpcExceptionFilter } from '@nestjs/microservices';
import { Observable, throwError } from 'rxjs';

@Catch(HttpException)
export class HttpToRpcExceptionFilter extends BaseRpcExceptionFilter {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  catch(exception: HttpException, _host: ArgumentsHost): Observable<never> {
    const status = exception.getStatus();
    const res = exception.getResponse();
    const message =
      typeof res === 'string'
        ? res
        : (res as { message?: string | string[] }).message;

    // Throw a plain object (not RpcException) so NATS serializes it flat:
    // { err: { statusCode, message } } — format expected by the gateway
    return throwError(() => ({
      statusCode: status,
      message: Array.isArray(message) ? message[0] : (message ?? res),
    }));
  }
}
