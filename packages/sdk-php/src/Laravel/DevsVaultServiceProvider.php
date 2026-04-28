<?php

declare(strict_types=1);

namespace DevsVault\Laravel;

use DevsVault\Client;
use Illuminate\Support\ServiceProvider;

final class DevsVaultServiceProvider extends ServiceProvider
{
    public function register(): void
    {
        $this->app->singleton(Client::class, function (): Client {
            $config = $this->app->make('config');

            return new Client(
                (string) $config->get('services.devsvault.url'),
                (string) $config->get('services.devsvault.token'),
            );
        });
    }
}