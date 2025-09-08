// pages/api/resumes.js
export const config = {
    api: {
        bodyParser: false,
    },
};

export default async function handler(req, res) {
    if (req.method !== 'POST') {
        return res.status(405).json({ error: 'Method not allowed' });
    }

    const token = req.headers.authorization;

    if (!token) {
        return res.status(401).json({ error: 'Unauthorized' });
    }

    try {
        // Проксируем запрос к interview service
        const response = await fetch('http://localhost:8081/api/resumes', {
            method: 'POST',
            headers: {
                'Authorization': token,
            },
            body: req,
        });

        const data = await response.json();

        if (response.ok) {
            res.status(201).json(data);
        } else {
            res.status(response.status).json(data);
        }
    } catch (error) {
        console.error('API Error:', error);
        res.status(500).json({ error: 'Server error' });
    }
}
