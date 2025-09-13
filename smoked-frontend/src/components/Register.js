import React, { useState } from 'react';
import api from '../utils/api';
import { useNavigate } from 'react-router-dom';

const Register = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [phone_number, setPhoneNumber] = useState('');
    const [name, setName] = useState('');
    const [error, setError] = useState('');
    const navigate = useNavigate();

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            await api.post('register', { username, password, phone_number, name});
            navigate('/login');
        } catch (err) {
            setError(err.response?.data?.error || 'Registration error');
        }
    };

    const Back = () => {
    navigate('/');
  };

    return (
        <form onSubmit={handleSubmit}>
            <div className="mb-3">
                <label>Login:</label>
                <input type="text" value={username} onChange={(e) => setUsername(e.target.value)} className="form-control" required />
            </div>
            <div className="mb-3">
                <label>Password:</label>
                <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} className="form-control" minLength="6" required />
            </div>
            <div className="mb-3">
                <label>Phone number:</label>
                <input type="tel" value={phone_number} onChange={(e) => setPhoneNumber(e.target.value)} className="form-control" placeholder="+7 (978) 123-45-67" required />
            </div>
            <div className="mb-3">
                <label>Name (Not necessarily):</label>
                <input type="text" value={name} onChange={(e) => setName(e.target.value)} className="form-control" placeholder="Your name" />
            </div>
            {error && <div className="alert alert-danger">{error}</div>}
            <button type="submit" className="btn btn-primary">Register</button>
            <button onClick={Back} className="btn btn-primary">Back</button>
        </form>
    );
};

export default Register;