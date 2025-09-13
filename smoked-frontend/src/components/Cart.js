import React, { useContext, useState, useEffect } from 'react';
import api from '../utils/api';
import { CartContext } from '../context/CartContext';
import './AssortmentList.css';  
import { useNavigate } from 'react-router-dom';

    const Cart = ({ isAuthenticated }) => {
    const { cart, removeFromCart, updateQuantity, updateSpice, clearCart, getTotalPrice } = useContext(CartContext);
    const [totalPrice, setTotalPrice] = useState(0);
    const [showModal, setShowModal] = useState(false);
    const [orderDetails, setOrderDetails] = useState(null);
    const navigate = useNavigate();
    const [error, setError] = useState(null);

    useEffect(() => {
        const calculateTotal = async () => {
            try {
                const items = cart.map((item) => ({
                    id: item.id,
                    quantity: item.quantity,
                    selectedSpice: item.selectedSpice || item.spice.recipe1,
                }));
                const response = await api.post('/calculate-bulk', { items });
                setTotalPrice(response.data.total_price);
            } catch (error) {
                console.error('Error to calculate total price:', error);
                setTotalPrice(getTotalPrice());
            }
        };
        if (cart.length > 0) {
      calculateTotal();
    } else {
      setTotalPrice(0);
    }
  }, [cart, getTotalPrice]);

  useEffect(() => {
    console.log('showModal:', showModal, 'orderDetails:', orderDetails); 
  }, [showModal, orderDetails]);

    const handleQuantityChange = (itemKey, value) => {
        updateQuantity(itemKey, parseFloat(value) || 1);
    };

    const handleSpiceChange = (item, value) => {
        if (value) { 
      updateSpice(item.itemKey, value);
    };
    };

    if (error) {
        return <div className="alert alert-danger">{error}</div>;
    }

    const handleOrder = async () => {
        if (cart.length === 0) {
            alert('Корзина пуста!');
            return;
        }

        try {
      const items = cart.map((item) => ({
        id: item.id,
        quantity: item.quantity,
        selected_spice: item.selectedSpice,
        meat: item.meat,
      }));
      console.log('Sending order data:', JSON.stringify({ items }, null, 2));
      const response = await api.post('/order', { items });
      console.log('Server response:', response.data);
      const newOrderDetails = {
        order_id: response.data.order_id,
        created_at: new Date(response.data.created_at).toLocaleString(),
        items: Array.isArray(response.data.items) ? response.data.items : [],
        total_price: response.data.total_price || 0,
      };
      setOrderDetails(newOrderDetails);
      setShowModal(true);
      clearCart();
    } catch (error) {
      console.error('Error to create order:', error.response?.data || error.message);
      setError('Failed to create order: ' + (error.response?.data?.error || error.message));
    }
  };

  const closeModal = () => {
    setShowModal(false);
    setOrderDetails(null);
    navigate('/');
  };

  const Back = () => {
    navigate('/');
  };

    return (
        <div>
            <h2>My cart</h2>
            {cart.length === 0 ? (
                <div>
                    <p style={{ color: '#FFFF00', textShadow: '1px 1px 2px rgba(0, 0, 0, 0.8)' }}>Your cart is empty</p>
                    <button onClick={Back} className="btn btn-primary">
                            Back
                        </button>
                </div>
            ) : (
                <>
                    <table className="table table-striped table-transparent">
                        <thead>
                            <tr>
                        <th>Meat</th>
                        <th>Price</th>
                        <th>Quantity (kg)</th>
                        <th>Total Price</th>
                        <th>Spice</th>
                        <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {cart.map((item) => (
                                <tr key={item.id}>
                                    <td>{item.meat}</td>
                                    <td>{item.price} rub.</td>
                                    <td>
                                        <input
                                            type="number"
                                            min="1"
                                            value={item.quantity}
                                            onChange={(e) => handleQuantityChange(item.itemKey, e.target.value)}
                                            className="form-control"
                                            style={{
                                                width: '80px',
                                                backgroundColor: 'transparent',
                                                color: '#FFFF00',
                                                textShadow: '1px 1px 2px rgba(0, 0, 0, 0.8)',
                                            }}
                                        />
                                    </td>
                                    <td>{(item.price * item.quantity).toFixed(2)} rub.</td>
                                    <td>
                                        <select
                                            className="form-select"
                                            style={{ width: '120px' }}
                                            value={item.selectedSpice || item.spice.recipe1 || item.spice.recipe2}
                                            onChange={(e) => handleSpiceChange(item, e.target.value)}
                                        >
                      {item.spice.recipe1 && (
                        <option value={item.spice.recipe1}>{item.spice.recipe1}</option>
                      )}
                      {item.spice.recipe2 && (
                        <option value={item.spice.recipe2}>{item.spice.recipe2}</option>
                      )}
                                        </select>
                                    </td>
                                    <td>
                                        <button
                                            onClick={() => removeFromCart(item.itemKey)}
                                            className="btn btn-sm btn-danger"
                                        >
                                            Delete
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                    <div className="mt-3">
                        <h4>
                            Price for all products: {totalPrice.toFixed(2)} rub.
                        </h4>
                        <h5 style={{ color: '#fc0808ff', textShadow: '1px 1px 2px rgba(0, 0, 0, 0.8)' }}>
                            Get 8% discount when you buy 10kg+! Or get 12% off when buying 20kg+!
                        </h5>
                        <button onClick={Back} className="btn btn-primary">
                            Back
                        </button>
                        <button onClick={clearCart} className="btn btn-danger">
                            Clear the cart
                        </button>
                        {isAuthenticated ? (
                        <button onClick={handleOrder} className="btn btn-primary">
                            Order
                        </button>
                        ) : (
                          <h3>
                            You need to register before make order, it's fast. Try it!
                          </h3>
                        )}
                    </div>
                </>
            )}
            {showModal && (
        <div className="modal">
            {console.log('Модал рендерится')}
          <div className="modal-content">
            <h3 className="modal-content-h3">Order accepted</h3>
            {orderDetails ? (
              <>
                <p><strong>Order number:</strong> {orderDetails.order_id}</p>
                <p><strong>Date:</strong> {orderDetails.created_at}</p>
                <h4 className="modal-content-h4">Products:</h4>
                <ul>
                  {orderDetails.items && orderDetails.items.length > 0 ? (
                    orderDetails.items.map((item, index) => (
                      <li key={index}>
                        Meat: {item.meat}: {item.quantity} kg, Spice: {item.selected_spice}
                      </li>
                    ))
                  ) : (
                    <li>No products</li>
                  )}
                </ul>
                <p><strong>Total price:</strong> {orderDetails.total_price.toFixed(2)} rub.</p>
              </>
            ) : (
              <p>Loading of order data...</p>
            )}
            <button onClick={closeModal} className="btn btn-primary">
              Close
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Cart;