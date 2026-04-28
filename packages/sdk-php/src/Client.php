<?php

declare(strict_types=1);

namespace DevsVault;

final class Client
{
    public function __construct(
        private readonly string $baseUrl,
        private readonly string $token,
    ) {
        if ($baseUrl === '' || $token === '') {
            throw new \InvalidArgumentException('baseUrl and token are required');
        }
    }

    public function getSecret(string $path): string
    {
        $resolved = $this->resolve($path);

        return $resolved['value'];
    }

    /**
     * @return array{path:string,version:int,value:string}
     */
    public function resolve(string $path): array
    {
        if (count(explode('/', $path)) !== 4) {
            throw new \InvalidArgumentException('path must be workspace/project/environment/name');
        }

        $url = rtrim($this->baseUrl, '/') . '/api/v1/secrets/resolve?path=' . rawurlencode($path);
        $context = stream_context_create([
            'http' => [
                'method' => 'GET',
                'header' => "Authorization: Bearer {$this->token}\r\nAccept: application/json\r\n",
                'ignore_errors' => true,
                'timeout' => 10,
            ],
        ]);

        $body = file_get_contents($url, false, $context);
        if ($body === false) {
            throw new \RuntimeException('secret resolve failed');
        }

        $status = $this->statusCode($http_response_header ?? []);
        if ($status < 200 || $status >= 300) {
            throw new \RuntimeException("secret resolve failed with status {$status}");
        }

        $decoded = json_decode($body, true, flags: JSON_THROW_ON_ERROR);
        if (!is_array($decoded) || !isset($decoded['value'])) {
            throw new \RuntimeException('invalid secret response');
        }

        return $decoded;
    }

    /** @param list<string> $headers */
    private function statusCode(array $headers): int
    {
        $line = $headers[0] ?? '';
        if (preg_match('/\s(\d{3})\s/', $line, $matches) !== 1) {
            return 0;
        }

        return (int) $matches[1];
    }
}