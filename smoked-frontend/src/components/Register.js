import React, { useState } from 'react';
import api from '../utils/api';
import { useNavigate } from 'react-router-dom';

const Register = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const navigate = useNavigate();

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            await api.post('register', { username, password });
            navigate('/login');
        } catch (err) {
            setError(err.response?.data?.error || 'Ошибка регистрации');
        }
    };

    return (
        <form onSubmit={handleSubmit}>
            <div className="mb-3">
                <label>Login:</label>
                <input type="text" value={username} onChange={(e) => setUsername(e.target.value)} className="form-control" required />
            </div>
            <div className="mb-3">
                <label>Password:</label>
                <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} className="form-control" required />
            </div>
            {error && <div className="alert alert-danger">{error}</div>}
            <button type="submit" className="btn btn-primary">Register</button>
        </form>
    );
};

export default Register;