import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import DefaultLayout from '../layouts/DefaultLayout/DefaultLayout';
import InboxScreen from '../screens/Inbox/Inbox';
import Tickets from '../screens/Tickets/Tickets';
import TicketDetail from '../screens/TicketDetail/TicketDetail';
import Customers from '../screens/Customers/Customers';
import KnowledgeBase from '../screens/KnowledgeBase/KnowledgeBase';
import ArticleEditor from '../screens/ArticleEditor/ArticleEditor';
import CannedResponses from '../screens/CannedResponses/CannedResponses';
import Settings from '../screens/Settings/Settings';

const AppRoutes: React.FC = () => {
  return (
    <DefaultLayout>
      <Routes>
        <Route path="/" element={<InboxScreen />} />
        <Route path="/tickets" element={<Tickets />} />
        <Route path="/tickets/:id" element={<TicketDetail />} />
        <Route path="/customers" element={<Customers />} />
        <Route path="/knowledge" element={<KnowledgeBase />} />
        <Route path="/knowledge/new" element={<ArticleEditor />} />
        <Route path="/knowledge/:id/edit" element={<ArticleEditor />} />
        <Route path="/canned" element={<CannedResponses />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </DefaultLayout>
  );
};

export default AppRoutes;
