import { Axios } from 'axios';
import { PagedImages, ReducedImage } from './resources';
import { Config } from '../config/config';

export const createImage = async (axios: Axios, image: File): Promise<ReducedImage> => {
  const data = new FormData();
  data.set('file', image);
  const response = await axios.post(`/images`, data, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });

  return response.data as ReducedImage;
};

export const fetchImageData = async (axios: Axios, id: string): Promise<File> => {
  const response = await axios.get(`/images/${id}`);
  return response.data as File;
};

export const getImages = async (
  axios: Axios,
  currentPage: number,
  perPage: number,
  onlyUnassigned: boolean
) => {
  const response = await axios.get(
    `/images?page=${currentPage}&perPage=${perPage}&onlyUnassigned=${onlyUnassigned}`,
    {
      headers: {
        'Content-Type': 'application/json',
      },
    }
  );

  if (response.status != 200) {
    throw `Got error ${response}`;
  }

  const images = response.data as PagedImages;
  return images;
};

export const deleteImage = async (axios: Axios, id: string): Promise<ReducedImage> => {
  const response = await axios.delete(`/images/${id}`);
  return response.data as ReducedImage;
};

export const modifyImage = async (
  axios: Axios,
  id: string,
  rotation: number
): Promise<ReducedImage> => {
  const data = {
    rotation,
  };
  const response = await axios.patch(`/images/${id}`, data);
  return response.data as ReducedImage;
};

export const urlForImage = (config: Config, hash: string, width: number | undefined) => {
  if (width !== undefined) {
    return `${config.apiHost}/assets/${hash}?width=${width}`;
  } else {
    return `${config.apiHost}/assets/${hash}`;
  }
};
