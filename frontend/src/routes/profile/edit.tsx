import { useContext, useState } from 'react';
import { useNavigate } from 'react-router';
import { AuthContext } from '../../context/auth';
import { AxiosContext } from '../../context/axios';
import { EditableProfile, ProfileEditor } from '../../components/profile_editor';
import { patchProfile, ProfileUpdateParams } from '../../api/profile';
import { YellowButton } from '../../components/shared';
import { createImage, modifyImage } from '../../api/image';

export const EditProfile = () => {
  const axiosInstance = useContext(AxiosContext);
  const navigate = useNavigate();
  const authContext = useContext(AuthContext);
  const profile = authContext.profile;

  const [editedData, setEditedData] = useState<null | EditableProfile>(null);

  const update = async () => {
    if (!axiosInstance || !editedData) {
      return;
    }

    const editedImage = editedData.image;
    let imageId = null;
    let targetRotation = 0;
    if (editedImage) {
      if (editedImage.type === 'url') {
        imageId = editedImage.image.id;
        targetRotation = editedImage.rotation;
      } else {
        const image = await createImage(axiosInstance, editedImage.file);
        imageId = image.id;
        targetRotation = editedImage.rotation;
      }
    }

    if (targetRotation !== 0 && imageId) {
      await modifyImage(axiosInstance, imageId, targetRotation);
    }

    await patchProfile(axiosInstance, {
      name: editedData.name,
      fullName: editedData.fullName,
      information: editedData.information,
      imageId,
    } as ProfileUpdateParams);
    console.log('Updated profile');
    authContext.invalidateProfile();
    navigate('/user/profile');
  };

  if (!profile) {
    return <div>No Profile</div>;
  }

  return (
    <ProfileEditor profile={profile} onUpdateProfile={setEditedData}>
      <YellowButton onClick={update}>Update Profile</YellowButton>
    </ProfileEditor>
  );
};
