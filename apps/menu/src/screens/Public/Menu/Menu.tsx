import { useParams } from 'react-router-dom';
import Public from '../../../layouts/Public/public';

const Menu = () => {
  const { id } = useParams<{ id: string }>();
  return <Public>{null}</Public>;
};

export default Menu;
