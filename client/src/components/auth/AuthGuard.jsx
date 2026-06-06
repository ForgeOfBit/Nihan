import { Navigate } from 'react-router-dom';
import useAuthStore from '../../stores/authStore';

export default function AuthGuard({ children }) {
  const { user, isAuthenticated } = useAuthStore();

  if (!isAuthenticated || !user) {
    return <Navigate to="/login" replace />;
  }

  return children;
}
