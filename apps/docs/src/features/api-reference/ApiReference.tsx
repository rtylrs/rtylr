import React from 'react';
import { HelmetContainer } from '@voxire/components';

interface Endpoint {
  method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  path: string;
  description: string;
}

interface EndpointGroup {
  name: string;
  endpoints: Endpoint[];
}

const endpointGroups: EndpointGroup[] = [
  {
    name: 'Authentication',
    endpoints: [
      {
        method: 'POST',
        path: '/auth/login',
        description: 'Authenticate and receive an access token',
      },
      { method: 'POST', path: '/auth/refresh', description: 'Refresh an expired access token' },
      { method: 'POST', path: '/auth/logout', description: 'Revoke the current access token' },
    ],
  },
  {
    name: 'Users',
    endpoints: [
      { method: 'GET', path: '/users', description: 'List all users in the organization' },
      { method: 'GET', path: '/users/:id', description: 'Get a specific user by ID' },
      { method: 'POST', path: '/users', description: 'Create a new user' },
      { method: 'PATCH', path: '/users/:id', description: 'Update user details' },
      { method: 'DELETE', path: '/users/:id', description: 'Remove a user from the organization' },
    ],
  },
  {
    name: 'Products',
    endpoints: [
      { method: 'GET', path: '/erp/products', description: 'List all ERP products' },
      { method: 'GET', path: '/erp/products/:id', description: 'Get a specific ERP product' },
      { method: 'POST', path: '/erp/products', description: 'Create a new ERP product' },
      { method: 'PUT', path: '/erp/products/:id', description: 'Replace an ERP product' },
      { method: 'DELETE', path: '/erp/products/:id', description: 'Delete an ERP product' },
    ],
  },
  {
    name: 'Orders',
    endpoints: [
      { method: 'GET', path: '/orders', description: 'List all orders' },
      { method: 'GET', path: '/orders/:id', description: 'Get a specific order' },
      { method: 'POST', path: '/orders', description: 'Create a new order' },
      { method: 'PATCH', path: '/orders/:id', description: 'Update order status' },
    ],
  },
  {
    name: 'Customers',
    endpoints: [
      { method: 'GET', path: '/customers', description: 'List all customers' },
      { method: 'GET', path: '/customers/:id', description: 'Get a specific customer' },
      { method: 'POST', path: '/customers', description: 'Create a new customer' },
      { method: 'PATCH', path: '/customers/:id', description: 'Update customer details' },
    ],
  },
  {
    name: 'Webhooks',
    endpoints: [
      { method: 'GET', path: '/webhooks', description: 'List all registered webhooks' },
      { method: 'POST', path: '/webhooks', description: 'Register a new webhook' },
      { method: 'DELETE', path: '/webhooks/:id', description: 'Delete a webhook' },
    ],
  },
];

const methodColors: Record<string, { bg: string; text: string }> = {
  GET: { bg: '#ecfdf5', text: '#059669' },
  POST: { bg: '#eff6ff', text: '#2563eb' },
  PUT: { bg: '#fefce8', text: '#ca8a04' },
  PATCH: { bg: '#fff7ed', text: '#ea580c' },
  DELETE: { bg: '#fef2f2', text: '#dc2626' },
};

const ApiReference: React.FC = () => {
  return (
    <>
      <HelmetContainer title="API reference · rtylr" />
      <div>
        <h1 className="text-2xl font-bold mb-2" style={{ color: 'var(--docs-text)' }}>
          API Reference
        </h1>
        <p className="text-sm mb-8" style={{ color: 'var(--docs-text-muted)' }}>
          The rtylr REST API. Base URL:{' '}
          <code
            className="text-xs px-1.5 py-0.5 rounded"
            style={{ background: 'var(--docs-surface)', border: '1px solid var(--docs-border)' }}
          >
            https://api.rtylr.com/v1
          </code>
        </p>

        {endpointGroups.map((group) => (
          <div key={group.name} className="mb-8">
            <h2 className="text-lg font-semibold mb-3" style={{ color: 'var(--docs-text)' }}>
              {group.name}
            </h2>
            <div
              className="border rounded-lg overflow-hidden"
              style={{ borderColor: 'var(--docs-border)' }}
            >
              {group.endpoints.map((endpoint, idx) => {
                const colors = methodColors[endpoint.method];
                return (
                  <div
                    key={`${endpoint.method}-${endpoint.path}`}
                    className="flex items-center gap-3 px-4 py-2.5 text-sm"
                    style={{
                      borderTop: idx > 0 ? `1px solid var(--docs-border)` : 'none',
                      background: 'var(--docs-bg)',
                    }}
                  >
                    <span
                      className="text-xs font-mono font-semibold px-2 py-0.5 rounded w-16 text-center flex-shrink-0"
                      style={{ background: colors.bg, color: colors.text }}
                    >
                      {endpoint.method}
                    </span>
                    <span className="font-mono text-sm" style={{ color: 'var(--docs-text)' }}>
                      {endpoint.path}
                    </span>
                    <span className="ml-auto text-sm" style={{ color: 'var(--docs-text-muted)' }}>
                      {endpoint.description}
                    </span>
                  </div>
                );
              })}
            </div>
          </div>
        ))}
      </div>
    </>
  );
};

export default ApiReference;
