import React, { useState, useEffect } from 'react';
import { timerAPI } from '../api';

const WeekHistory = () => {
  const [weeks, setWeeks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const formatDate = (dateString) => {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const formatDuration = (minutes) => {
    if (!minutes) return '0:00';
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    return `${hours}:${mins.toString().padStart(2, '0')}`;
  };

  const getStatus = (week) => {
    if (!week.started_at) return 'Не начата';
    if (week.ended_at) return 'Завершена';
    if (week.is_paused) return 'Приостановлена';
    return 'Активна';
  };

  const getStatusClass = (week) => {
    if (!week.started_at) return 'status-not-started';
    if (week.ended_at) return 'status-completed';
    if (week.is_paused) return 'status-paused';
    return 'status-active';
  };

  const getProgressInfo = (week) => {
    if (!week.week_goal_minutes || week.week_goal_minutes <= 0) {
      return { percentage: 0, text: 'Без цели' };
    }
    
    const percentage = Math.round((week.total_work_minutes / week.week_goal_minutes) * 100);
    const goalText = formatDuration(week.week_goal_minutes);
    return { 
      percentage: Math.min(percentage, 100), 
      text: `${percentage}% из ${goalText}`,
      achieved: percentage >= 100
    };
  };

  useEffect(() => {
    fetchWeekHistory();
  }, []);

  const fetchWeekHistory = async () => {
    try {
      setLoading(true);
      const response = await timerAPI.getWeekHistory();
      setWeeks(response.data || []);
    } catch (err) {
      setError('Ошибка загрузки истории недель');
      console.error('Error fetching week history:', err);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="container">
        <div className="loading">Загрузка истории недель...</div>
      </div>
    );
  }

  return (
    <div className="container">
      <div className="header">
        <h1>История рабочих недель</h1>
        <button onClick={fetchWeekHistory} className="btn btn-secondary">
          Обновить
        </button>
      </div>

      {error && <div className="error">{error}</div>}

      <div className="week-history">
        {weeks.length === 0 ? (
          <div className="no-data">
            <p>История недель пуста</p>
          </div>
        ) : (
          <div className="table-container">
            <table className="weeks-table">
              <thead>
                <tr>
                  <th>№ Недели</th>
                  <th>Начало</th>
                  <th>Конец</th>
                  <th>Статус</th>
                  <th>Продолжительность</th>
                  <th>Цель / Прогресс</th>
                </tr>
              </thead>
              <tbody>
                {weeks.map((week) => {
                  const progress = getProgressInfo(week);
                  return (
                    <tr key={week.id}>
                      <td>{week.week_number || week.id}</td>
                      <td>{formatDate(week.started_at)}</td>
                      <td>{formatDate(week.ended_at)}</td>
                      <td>
                        <span className={`status ${getStatusClass(week)}`}>
                          {getStatus(week)}
                        </span>
                      </td>
                      <td>{formatDuration(week.total_work_minutes)}</td>
                      <td>
                        <div className="progress-cell">
                          <span className={progress.achieved ? 'progress-achieved' : 'progress-text'}>
                            {progress.text}
                          </span>
                          {week.week_goal_minutes > 0 && (
                            <div className="mini-progress-bar">
                              <div 
                                className="mini-progress-fill"
                                style={{ 
                                  width: `${progress.percentage}%`,
                                  backgroundColor: progress.achieved ? '#28a745' : '#007bff'
                                }}
                              ></div>
                            </div>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};

export default WeekHistory;
