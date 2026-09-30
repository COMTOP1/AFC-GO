import { lazy } from 'react';
import { Route, Routes } from 'react-router';

import Layout from './components/layout/Layout';
import HomePage from './pages/home/HomePage';

const TeamsPage = lazy(() => import('./pages/teams/TeamsPage'));
const TeamPage = lazy(() => import('./pages/teams/TeamPage'));
const NewsListPage = lazy(() => import('./pages/news/NewsListPage'));
const NewsArticlePage = lazy(() => import('./pages/news/NewsArticlePage'));
const NewsFormPage = lazy(() => import('./pages/news/NewsFormPage'));
const WhatsOnPage = lazy(() => import('./pages/whatson/WhatsOnPage'));
const EventPage = lazy(() => import('./pages/whatson/EventPage'));
const EventFormPage = lazy(() => import('./pages/whatson/EventFormPage'));
const GalleryPage = lazy(() => import('./pages/gallery/GalleryPage'));
const DocumentsPage = lazy(() => import('./pages/documents/DocumentsPage'));
const ProgrammesPage = lazy(() => import('./pages/programmes/ProgrammesPage'));
const SponsorsPage = lazy(() => import('./pages/sponsors/SponsorsPage'));
const InfoPage = lazy(() => import('./pages/info/InfoPage'));
const ContactPage = lazy(() => import('./pages/contact/ContactPage'));
const AccountPage = lazy(() => import('./pages/account/AccountPage'));
const ResetPage = lazy(() => import('./pages/reset/ResetPage'));
const DesignPage = lazy(() => import('./pages/DesignPage'));
const NotFoundPage = lazy(() => import('./pages/NotFoundPage'));

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<HomePage />} />
        <Route path="teams" element={<TeamsPage />} />
        <Route path="team/:id" element={<TeamPage />} />
        <Route path="news" element={<NewsListPage />} />
        <Route path="news/new" element={<NewsFormPage />} />
        <Route path="news/:id/edit" element={<NewsFormPage />} />
        <Route path="news/:id" element={<NewsArticlePage />} />
        <Route path="whatson" element={<WhatsOnPage />} />
        <Route path="whatson/new" element={<EventFormPage />} />
        <Route path="whatson/:id/edit" element={<EventFormPage />} />
        <Route path="whatson/:id" element={<EventPage />} />
        <Route path="gallery" element={<GalleryPage />} />
        <Route path="documents" element={<DocumentsPage />} />
        <Route path="programmes" element={<ProgrammesPage />} />
        <Route path="sponsors" element={<SponsorsPage />} />
        <Route path="info" element={<InfoPage />} />
        <Route path="contact" element={<ContactPage />} />
        <Route path="account" element={<AccountPage />} />
        <Route path="reset/:token" element={<ResetPage />} />
        <Route path="design" element={<DesignPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  );
}
