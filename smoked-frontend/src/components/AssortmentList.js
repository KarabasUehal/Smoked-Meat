import React, { useState, useEffect, useContext} from 'react';
import api from '../utils/api';
import { Link } from 'react-router-dom';
import './AssortmentList.css';
import { CartContext } from '../context/CartContext';

const AssortmentList = ({ isAuthenticated }) => {
    const [products, setProducts] = useState([]);
    const [error, setError] = useState(''); 
    const [quantities, setQuantities] = useState({});
    const [selectedSpices, setSelectedSpices] = useState({});
    const { addToCart } = useContext(CartContext);

    useEffect(() => {
        fetchProducts();
    }, []);

    const fetchProducts = async () => {
        try {
            const response = await api.get('assortment');
            setProducts(response.data);
            setQuantities(
                response.data.reduce((acc, product) => ({
                    ...acc,
                    [product.id]: 1,
                }), {})
            );
            setSelectedSpices(
                response.data.reduce((acc, product) => ({
                    ...acc,
                    [product.id]: product.spice.recipe1 || product.spice.recipe2 || '',
                }), {})
            );
        } catch (error) {
            console.error('Ошибка при загрузке ассортимента:', error);
        }
    };

    const deleteProduct = async (id) => {
        if (window.confirm('Удалить продукт?')) {
            try {
                await api.delete(`/product/${id}`);
                fetchProducts();
            } catch (error) {
                console.error('Ошибка при удалении:', error);
            }
        }
    };

    const handleQuantityChange = (id, value) => {
        setQuantities((prev) => ({
            ...prev,
            [id]: Math.max(1, parseFloat(value) || 1),
        }));
    };

    const handleSpiceChange = (id, value) => {
        setSelectedSpices((prev) => ({
            ...prev,
            [id]: value,
        }));
    };

    const handleAddToCart = (product) => {
        const quantity = quantities[product.id] || 1;
        const selectedSpice = selectedSpices[product.id] || product.spice.recipe1 || product.spice.recipe2;
        addToCart(product, quantity, selectedSpice);
    };

     if (error) {
        return <div className="alert alert-danger">{error}</div>;
    }

    if (products.length === 0 && !error) {
        return <div>Loading...</div>;
    }

    return (
        <div>
            {isAuthenticated && (
                <Link to="/add" className="btn btn-warning mb-3">
                    Add product
                </Link>
            )}
                <table className="table table-striped table-transparent">
                    <thead>
                    <tr>{isAuthenticated ? (
                        <th>ID</th>
                            ) : (
                         <th style={{textShadow: '2px 2px 4px rgba(255, 0, 0, 0.8)' }}>Hot propose!</th>
                                            )}
                        <th>Meat</th>
                        <th>Availability</th>
                        <th>Price</th>
                        <th>Spice</th>
                        <th>Quantity (kg)</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {products.map((product) => (
                        <tr key={product.id}>
                            {isAuthenticated ? (
                            <td>{product.id}</td>
                                            ) : (
                            <td style={{textShadow: '2px 2px 4px rgba(255, 0, 0, 0.8)' }}>New!</td>
                                            )}
                            <td>{product.meat}</td>
                            <td>{product.avail ? 'Yes' : 'No'}</td>
                            <td>{product.price} rub.</td>
                            <td>
                                {product.spice.recipe1 && `${product.spice.recipe1} /`}
                                <br></br>
                                {product.spice.recipe2 && `${product.spice.recipe2}`}
                            </td>
                            <td>
                                <div className="d-flex align-items-center">
                                    <select
                                        className="form-select me-2"
                                        style={{ width: '120px' }}
                                        value={selectedSpices[product.id] || product.spice.recipe1 || product.spice.recipe2}
                                        onChange={(e) => handleSpiceChange(product.id, e.target.value)}
                                    >
                                        {product.spice.recipe1 && <option className="select-option" value={product.spice.recipe1}>{product.spice.recipe1}</option>}
                                        {product.spice.recipe2 && <option className="select-option" value={product.spice.recipe2}>{product.spice.recipe2}</option>}
                                    </select>
                                    <input
                                        type="number"
                                        min="1"
                                        value={quantities[product.id] || 1}
                                        onChange={(e) => handleQuantityChange(product.id, e.target.value)}
                                        className="form-control"
                                        style={{
                                            width: '80px',
                                            backgroundColor: 'transparent',
                                            color: '#FFFF00',
                                            textShadow: '1px 1px 2px rgba(0, 0, 0, 0.8)',
                                        }}
                                    />
                                </div>
                            </td>
                            <td>
                                <button
                                    onClick={() => handleAddToCart(product)}
                                    className="btn btn-sm btn-success me-2"
                                    disabled={!product.avail}
                                >
                                 Add to cart
                                </button>
                                {isAuthenticated && (
                                    <>
                                        <Link
                                            to={`/edit/${product.id}`}
                                            className="btn btn-sm btn-warning me-2"
                                        >
                                            Edit
                                        </Link>
                                        <button
                                            onClick={() => deleteProduct(product.id)}
                                            className="btn btn-sm btn-danger"
                                        >Delete
                                        </button>
                                    </>
                                )}
                            </td>
                        </tr>
                    ))}
                </tbody>
                </table>
        </div>
    );
};

export default AssortmentList;