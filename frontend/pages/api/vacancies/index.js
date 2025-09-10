// pages/api/vacancies/index.js
export default async function handler(req, res) {
    const token = req.headers.authorization;

    if (!token) {
        return res.status(401).json({ error: 'Unauthorized' });
    }

    if (req.method === 'GET') {
        try {
            const response = await fetch('http://localhost:8081/api/vacancies', {
                headers: {
                    'Authorization': token,
                },
            });

            const data = await response.json();

            if (response.ok) {
                res.json(data);
            } else {
                res.status(response.status).json(data);
            }
        } catch (error) {
            console.error('API Error:', error);
            res.status(500).json({ error: 'Server error' });
        }
    } else {
        res.setHeader('Allow', ['GET']);
        res.status(405).json({ error: 'Method not allowed' });
    }
}
