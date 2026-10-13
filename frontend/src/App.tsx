import { Navigate, Outlet, Route, Routes } from "react-router-dom";
import { isLoggedIn } from "./stores/auth";
import Layout from "./components/Layout";
import LoginPage from "./pages/LoginPage";
import RegisterPage from "./pages/RegisterPage";
import HomePage from "./pages/HomePage";
import CustomSetupPage from "./pages/CustomSetupPage";
import CampaignPage from "./pages/CampaignPage";
import GamePage from "./pages/GamePage";
import ResultPage from "./pages/ResultPage";
import DecksPage from "./pages/DecksPage";
import DeckEditPage from "./pages/DeckEditPage";
import MarketplacePage from "./pages/MarketplacePage";
import MarketDeckPage from "./pages/MarketDeckPage";
import StatsPage from "./pages/StatsPage";
import MergePage from "./pages/MergePage";

function Protected() {
  if (!isLoggedIn()) return <Navigate to="/login" replace />;
  return (
    <Layout>
      <Outlet />
    </Layout>
  );
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route element={<Protected />}>
        <Route path="/" element={<HomePage />} />
        <Route path="/play" element={<CustomSetupPage />} />
        <Route path="/campaign" element={<CampaignPage />} />
        <Route path="/game" element={<GamePage />} />
        <Route path="/result" element={<ResultPage />} />
        <Route path="/decks" element={<DecksPage />} />
        <Route path="/decks/new" element={<DeckEditPage />} />
        <Route path="/decks/:id" element={<DeckEditPage />} />
        <Route path="/decks/:id/merge" element={<MergePage />} />
        <Route path="/market" element={<MarketplacePage />} />
        <Route path="/market/:id" element={<MarketDeckPage />} />
        <Route path="/stats" element={<StatsPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
