import React, { useState, useEffect } from 'react';
import Auth from './components/Auth';
import Timer from './components/Timer';
import WeekHistory from './components/WeekHistory';
import './index.css';

function App() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState('timer');

  useEffect(() => {
    // Проверяем, есть ли сохраненный токен и данные пользователя
    const token = localStorage.getItem('token');
    const savedUser = localStorage.getItem('user');

    if (token && savedUser) {
      try {
        const userData = JSON.parse(savedUser);
        setUser(userData);
      } catch (err) {
        // Если данные повреждены, очищаем их
        localStorage.removeItem('token');
        localStorage.removeItem('user');
      }
    }
    
    setLoading(false);
  }, []);

  const handleLogin = (userData) => {
    setUser(userData);
    setCurrentPage('timer');
  };

  const handleLogout = () => {
    setUser(null);
    setCurrentPage('timer');
  };

  const renderCurrentPage = () => {
    switch (currentPage) {
      case 'history':
        return <WeekHistory />;
      case 'timer':
      default:
        return <Timer user={user} onLogout={handleLogout} />;
    }
  };

  if (loading) {
    return (
      <div className="container">
        <div style={{ textAlign: 'center', padding: '50px' }}>
          Загрузка...
        </div>
      </div>
    );
  }

  return (
    <div className="App">
      {user ? (
        <>
          <div className="navigation">
            <button 
              className={`nav-btn ${currentPage === 'timer' ? 'active' : ''}`}
              onClick={() => setCurrentPage('timer')}
            >
              Таймер
            </button>
            <button 
              className={`nav-btn ${currentPage === 'history' ? 'active' : ''}`}
              onClick={() => setCurrentPage('history')}
            >
              История недель
            </button>
            <button 
              className="nav-btn logout-btn"
              onClick={handleLogout}
              style={{ marginLeft: 'auto' }}
            >
              Выйти
            </button>
          </div>
          {renderCurrentPage()}
        </>
      ) : (
        <Auth onLogin={handleLogin} />
      )}
    </div>
  );
}

export default App;
